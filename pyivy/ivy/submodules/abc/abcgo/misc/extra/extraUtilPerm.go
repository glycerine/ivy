package extra

import (
	"fmt"
	"slices"
	"time"
)

/**CFile****************************************************************

  FileName    [extraUtilPerm.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Permutation computation.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilPerm.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

// ZDDs encode families of variable sets. Terminal 0 is the empty family;
// terminal 1 is the family containing the empty set. True edges include Var.
type Abc_ZddOper int32

const (
	ABC_ZDD_OPER_NONE Abc_ZddOper = iota
	ABC_ZDD_OPER_DIFF
	ABC_ZDD_OPER_UNION
	ABC_ZDD_OPER_MIN_UNION
	ABC_ZDD_OPER_INTER
	ABC_ZDD_OPER_PERM
	ABC_ZDD_OPER_PERM_PROD
	ABC_ZDD_OPER_COF0
	ABC_ZDD_OPER_COF1
	ABC_ZDD_OPER_THRESH
	ABC_ZDD_OPER_DOT_PROD
	ABC_ZDD_OPER_DOT_PROD_6
	ABC_ZDD_OPER_INSERT
	ABC_ZDD_OPER_PATHS
	ABC_ZDD_OPER_NODES
	ABC_ZDD_OPER_ITE
)

type Abc_ZddObj struct {
	Var, True, False int32
	Mark             bool
	id               int32
}
type Abc_ZddObj_ = Abc_ZddObj
type Abc_ZddEnt struct{ Arg0, Arg1, Arg2, Res int32 }
type Abc_ZddEnt_ = Abc_ZddEnt
type zddKey struct{ a, b, c int32 }
type Abc_ZddMan struct {
	nVars, nObjs, nObjsAlloc, nPermSize int32
	pObjs                               []Abc_ZddObj
	unique                              map[zddKey]int32
	cache                               map[zddKey]int32
	counts                              map[int32]uint64
	pV2TI, pV2TJ, pT2V                  []int32
	nCacheLookups, nCacheMisses         uint64
}
type Abc_ZddMan_ = Abc_ZddMan

func Abc_ZddIthVar(v int32) int32 { return v + 2 }
func Abc_ZddHash(a, b, c int32) uint32 {
	return 12582917*uint32(a) + 4256249*uint32(b) + 741457*uint32(c)
}

// Return a value, not a pointer into storage that subsequent creation may grow.
func Abc_ZddNode(p *Abc_ZddMan, id int32) Abc_ZddObj { return p.pObjs[id] }
func Abc_ZddObjId(p *Abc_ZddMan, obj Abc_ZddObj) int32 {
	if obj.id < 0 || obj.id >= p.nObjs {
		panic("invalid ZDD object ID")
	}
	return obj.id
}
func Abc_ZddObjVar(p *Abc_ZddMan, id int32) int32 { return p.pObjs[id].Var }
func Abc_ZddSetVarIJ(p *Abc_ZddMan, i, j, v int32) {
	if i < 0 || i >= j || j >= p.nPermSize {
		panic("invalid transposition")
	}
	p.pT2V[i*p.nPermSize+j] = v
}
func Abc_ZddVarIJ(p *Abc_ZddMan, i, j int32) int32 {
	if i < 0 || i >= j || j >= p.nPermSize {
		panic("invalid transposition")
	}
	return p.pT2V[i*p.nPermSize+j]
}
func Abc_ZddVarsClash(p *Abc_ZddMan, a, b int32) int32 {
	if p.pV2TI[a] == p.pV2TI[b] || p.pV2TI[a] == p.pV2TJ[b] || p.pV2TJ[a] == p.pV2TI[b] || p.pV2TJ[a] == p.pV2TJ[b] {
		return 1
	}
	return 0
}
func Abc_ZddCacheLookup(p *Abc_ZddMan, a, b, c int32) int32 {
	p.nCacheLookups++
	if result, ok := p.cache[zddKey{a, b, c}]; ok {
		return result
	}
	return -1
}
func Abc_ZddCacheInsert(p *Abc_ZddMan, a, b, c, result int32) int32 {
	if result < 0 {
		panic("negative ZDD result")
	}
	// Bound memo storage independently of the number of recursive requests.
	// Clearing a full memo changes performance only, never the unique table.
	if len(p.cache) >= min(int(p.nObjsAlloc), 1<<20) {
		clear(p.cache)
	}
	p.cache[zddKey{a, b, c}] = result
	p.nCacheMisses++
	return result
}
func Abc_ZddUniqueLookup(p *Abc_ZddMan, v, yes, no int32) int32 { return p.unique[zddKey{v, yes, no}] }
func Abc_ZddUniqueCreate(p *Abc_ZddMan, v, yes, no int32) int32 {
	if v < 0 || v >= p.nVars || v >= p.pObjs[yes].Var || v >= p.pObjs[no].Var {
		panic("unordered ZDD node")
	}
	if yes == 0 {
		return no
	}
	key := zddKey{v, yes, no}
	if id, ok := p.unique[key]; ok {
		return id
	}
	if p.nObjs >= p.nObjsAlloc {
		panic("ZDD node limit exceeded")
	}
	id := p.nObjs
	p.nObjs++
	p.pObjs = append(p.pObjs, Abc_ZddObj{Var: v, True: yes, False: no, id: id})
	p.unique[key] = id
	return id
}
func Abc_ZddBuildSet(p *Abc_ZddMan, set []int32, size int32) int32 {
	set = set[:size]
	slices.Sort(set)
	result := int32(1)
	for i := len(set) - 1; i >= 0; i-- {
		if i+1 < len(set) && set[i] == set[i+1] {
			continue
		}
		result = Abc_ZddUniqueCreate(p, set[i], result, 0)
	}
	return result
}
func Abc_ZddManAlloc(vars, limit int32) *Abc_ZddMan {
	if vars < 0 || limit < 2 || int64(vars)+2 > int64(limit) {
		panic("invalid ZDD manager dimensions")
	}
	p := &Abc_ZddMan{nVars: vars, nObjs: 2, nObjsAlloc: limit, pObjs: []Abc_ZddObj{{Var: vars, True: -1, False: -1, id: 0}, {Var: vars, True: -1, False: -1, id: 1}}, unique: map[zddKey]int32{}, cache: map[zddKey]int32{}, counts: map[int32]uint64{}}
	for v := int32(0); v < vars; v++ {
		Abc_ZddUniqueCreate(p, v, 1, 0)
	}
	return p
}
func Abc_ZddManCreatePerms(p *Abc_ZddMan, size int32) {
	if size < 0 || int64(size)*int64(size-1) != 2*int64(p.nVars) || p.nPermSize != 0 {
		panic("invalid ZDD permutation dimensions")
	}
	p.nPermSize = size
	p.pV2TI = make([]int32, p.nVars)
	p.pV2TJ = make([]int32, p.nVars)
	p.pT2V = make([]int32, int(size)*int(size))
	for i := range p.pT2V {
		p.pT2V[i] = -1
	}
	v := int32(0)
	for i := int32(0); i < size; i++ {
		for j := i + 1; j < size; j++ {
			p.pV2TI[v], p.pV2TJ[v] = i, j
			Abc_ZddSetVarIJ(p, i, j, v)
			v++
		}
	}
}
func Abc_ZddManFree(p *Abc_ZddMan) {
	if p == nil {
		return
	}
	p.pObjs = nil
	p.unique = nil
	p.cache = nil
	p.counts = nil
	p.pV2TI = nil
	p.pV2TJ = nil
	p.pT2V = nil
	p.nObjs = 0
}
func zddRemember(p *Abc_ZddMan, op Abc_ZddOper, a, b, result int32) int32 {
	return Abc_ZddCacheInsert(p, a, b, int32(op), result)
}
func zddRecall(p *Abc_ZddMan, op Abc_ZddOper, a, b int32) int32 {
	return Abc_ZddCacheLookup(p, a, b, int32(op))
}
func Abc_ZddDiff(p *Abc_ZddMan, a, b int32) int32 {
	if a == 0 || a == b {
		return 0
	}
	if b == 0 {
		return a
	}
	if r := zddRecall(p, ABC_ZDD_OPER_DIFF, a, b); r >= 0 {
		return r
	}
	A, B := p.pObjs[a], p.pObjs[b]
	var result int32
	if A.Var < B.Var {
		result = Abc_ZddUniqueCreate(p, A.Var, A.True, Abc_ZddDiff(p, A.False, b))
	} else if A.Var > B.Var {
		result = Abc_ZddDiff(p, a, B.False)
	} else {
		no := Abc_ZddDiff(p, A.False, B.False)
		yes := Abc_ZddDiff(p, A.True, B.True)
		result = Abc_ZddUniqueCreate(p, A.Var, yes, no)
	}
	return zddRemember(p, ABC_ZDD_OPER_DIFF, a, b, result)
}
func Abc_ZddUnion(p *Abc_ZddMan, a, b int32) int32    { return zddUnion(p, a, b, false) }
func Abc_ZddMinUnion(p *Abc_ZddMan, a, b int32) int32 { return zddUnion(p, a, b, true) }
func zddUnion(p *Abc_ZddMan, a, b int32, minimal bool) int32 {
	if a == 0 {
		return b
	}
	if b == 0 || a == b {
		return a
	}
	if a > b {
		a, b = b, a
	}
	op := ABC_ZDD_OPER_UNION
	if minimal {
		op = ABC_ZDD_OPER_MIN_UNION
	}
	if r := zddRecall(p, op, a, b); r >= 0 {
		return r
	}
	A, B := p.pObjs[a], p.pObjs[b]
	var yes, no int32
	if A.Var < B.Var {
		no = zddUnion(p, A.False, b, minimal)
		yes = A.True
	} else if A.Var > B.Var {
		no = zddUnion(p, a, B.False, minimal)
		yes = B.True
	} else {
		no = zddUnion(p, A.False, B.False, minimal)
		yes = zddUnion(p, A.True, B.True, minimal)
	}
	if minimal {
		yes = Abc_ZddDiff(p, yes, no)
	}
	return zddRemember(p, op, a, b, Abc_ZddUniqueCreate(p, min(A.Var, B.Var), yes, no))
}
func Abc_ZddIntersect(p *Abc_ZddMan, a, b int32) int32 {
	if a == 0 || b == 0 {
		return 0
	}
	if a == b {
		return a
	}
	if a > b {
		a, b = b, a
	}
	if r := zddRecall(p, ABC_ZDD_OPER_INTER, a, b); r >= 0 {
		return r
	}
	A, B := p.pObjs[a], p.pObjs[b]
	var r int32
	if A.Var < B.Var {
		r = Abc_ZddIntersect(p, A.False, b)
	} else if A.Var > B.Var {
		r = Abc_ZddIntersect(p, a, B.False)
	} else {
		no := Abc_ZddIntersect(p, A.False, B.False)
		yes := Abc_ZddIntersect(p, A.True, B.True)
		r = Abc_ZddUniqueCreate(p, A.Var, yes, no)
	}
	return zddRemember(p, ABC_ZDD_OPER_INTER, a, b, r)
}
func Abc_ZddCof0(p *Abc_ZddMan, a, v int32) int32 {
	if a < 2 || p.pObjs[a].Var > v {
		return a
	}
	if r := zddRecall(p, ABC_ZDD_OPER_COF0, a, v); r >= 0 {
		return r
	}
	A := p.pObjs[a]
	var r int32
	if A.Var < v {
		no := Abc_ZddCof0(p, A.False, v)
		yes := Abc_ZddCof0(p, A.True, v)
		r = Abc_ZddUniqueCreate(p, A.Var, yes, no)
	} else {
		r = Abc_ZddCof0(p, A.False, v)
	}
	return zddRemember(p, ABC_ZDD_OPER_COF0, a, v, r)
}

// Cof1 retains the variable on its true branch, as in the original API.
func Abc_ZddCof1(p *Abc_ZddMan, a, v int32) int32 {
	if a < 2 || p.pObjs[a].Var > v {
		return a
	}
	if r := zddRecall(p, ABC_ZDD_OPER_COF1, a, v); r >= 0 {
		return r
	}
	A := p.pObjs[a]
	var yes, no int32
	if A.Var < v {
		no = Abc_ZddCof1(p, A.False, v)
		yes = Abc_ZddCof1(p, A.True, v)
	} else {
		yes = Abc_ZddCof1(p, A.True, v)
	}
	return zddRemember(p, ABC_ZDD_OPER_COF1, a, v, Abc_ZddUniqueCreate(p, A.Var, yes, no))
}
func Abc_ZddCountPaths(p *Abc_ZddMan, a int32) uint64 {
	if a < 0 || a >= p.nObjs {
		panic("invalid ZDD root")
	}
	if a < 2 {
		return uint64(a)
	}
	if count, ok := p.counts[a]; ok {
		return count
	}
	A := p.pObjs[a]
	x, y := Abc_ZddCountPaths(p, A.False), Abc_ZddCountPaths(p, A.True)
	if ^uint64(0)-x < y {
		panic("ZDD path count overflow")
	}
	p.counts[a] = x + y
	return x + y
}
func Abc_ZddCount_rec(p *Abc_ZddMan, id int32) int32 {
	if id < 2 || p.pObjs[id].Mark {
		return 0
	}
	p.pObjs[id].Mark = true
	A := p.pObjs[id]
	return 1 + Abc_ZddCount_rec(p, A.False) + Abc_ZddCount_rec(p, A.True)
}
func Abc_ZddUnmark_rec(p *Abc_ZddMan, id int32) {
	if id < 2 || !p.pObjs[id].Mark {
		return
	}
	p.pObjs[id].Mark = false
	A := p.pObjs[id]
	Abc_ZddUnmark_rec(p, A.False)
	Abc_ZddUnmark_rec(p, A.True)
}
func Abc_ZddCountNodes(p *Abc_ZddMan, id int32) int32 {
	n := Abc_ZddCount_rec(p, id)
	Abc_ZddUnmark_rec(p, id)
	return n
}
func Abc_ZddCountNodesArray(p *Abc_ZddMan, nodes []int32) int32 {
	var n int32
	for _, id := range nodes {
		n += Abc_ZddCount_rec(p, id)
	}
	for _, id := range nodes {
		Abc_ZddUnmark_rec(p, id)
	}
	return n
}
func Abc_ZddThresh(p *Abc_ZddMan, a, b int32) int32 {
	if b < 0 {
		panic("negative ZDD threshold")
	}
	if a < 2 {
		return a
	}
	if b == 0 {
		// Only the empty member can survive a zero-size bound. Follow false
		// edges rather than dropping it when the family also has other sets.
		for a >= 2 {
			a = p.pObjs[a].False
		}
		return a
	}
	if r := zddRecall(p, ABC_ZDD_OPER_THRESH, a, b); r >= 0 {
		return r
	}
	A := p.pObjs[a]
	no := Abc_ZddThresh(p, A.False, b)
	yes := Abc_ZddThresh(p, A.True, b-1)
	return zddRemember(p, ABC_ZDD_OPER_THRESH, a, b, Abc_ZddUniqueCreate(p, A.Var, yes, no))
}
func Abc_ZddDotProduct(p *Abc_ZddMan, a, b int32) int32     { return zddDot(p, a, b, false) }
func Abc_ZddDotMinProduct6(p *Abc_ZddMan, a, b int32) int32 { return zddDot(p, a, b, true) }
func zddDot(p *Abc_ZddMan, a, b int32, minimal bool) int32 {
	if a == 0 || b == 0 {
		return 0
	}
	if a == 1 {
		return b
	}
	if b == 1 {
		return a
	}
	if a > b {
		a, b = b, a
	}
	op := ABC_ZDD_OPER_DOT_PROD
	if minimal {
		op = ABC_ZDD_OPER_DOT_PROD_6
	}
	if r := zddRecall(p, op, a, b); r >= 0 {
		return r
	}
	A, B := p.pObjs[a], p.pObjs[b]
	var yes, no int32
	if A.Var < B.Var {
		no = zddDot(p, A.False, b, minimal)
		yes = zddDot(p, A.True, b, minimal)
	} else if A.Var > B.Var {
		no = zddDot(p, a, B.False, minimal)
		yes = zddDot(p, a, B.True, minimal)
	} else {
		no = zddDot(p, A.False, B.False, minimal)
		tail := zddUnion(p, B.False, B.True, minimal)
		x := zddDot(p, A.True, tail, minimal)
		y := zddDot(p, A.False, B.True, minimal)
		yes = zddUnion(p, x, y, minimal)
	}
	if minimal {
		yes = Abc_ZddThresh(p, yes, 5)
		yes = Abc_ZddDiff(p, yes, no)
	}
	return zddRemember(p, op, a, b, Abc_ZddUniqueCreate(p, min(A.Var, B.Var), yes, no))
}
func Abc_ZddPerm(p *Abc_ZddMan, a, v int32) int32 {
	if v < 0 || v >= p.nVars || p.nPermSize == 0 {
		panic("invalid permutation variable")
	}
	if a == 0 {
		return 0
	}
	if a == 1 {
		return Abc_ZddIthVar(v)
	}
	if r := zddRecall(p, ABC_ZDD_OPER_PERM, a, v); r >= 0 {
		return r
	}
	A := p.pObjs[a]
	var r int32
	if p.pV2TI[A.Var] > p.pV2TI[v] {
		r = Abc_ZddUniqueCreate(p, v, a, 0)
	} else if A.Var == v {
		r = Abc_ZddUnion(p, Abc_ZddPerm(p, A.False, v), A.True)
	} else {
		ai, aj, bi, bj := p.pV2TI[A.Var], p.pV2TJ[A.Var], p.pV2TI[v], p.pV2TJ[v]
		var next, top int32
		switch {
		case aj == bi:
			next, top = v, Abc_ZddVarIJ(p, ai, bj)
		case aj == bj:
			next, top = v, Abc_ZddVarIJ(p, ai, bi)
		case ai == bi:
			next, top = Abc_ZddVarIJ(p, min(aj, bj), max(aj, bj)), A.Var
		default:
			next, top = v, A.Var
		}
		no := Abc_ZddPerm(p, A.False, v)
		yes := Abc_ZddPerm(p, A.True, next)
		if p.pObjs[no].Var > top {
			r = Abc_ZddUniqueCreate(p, top, yes, no)
		} else {
			r = Abc_ZddUnion(p, no, Abc_ZddUniqueCreate(p, top, yes, 0))
		}
	}
	return zddRemember(p, ABC_ZDD_OPER_PERM, a, v, r)
}
func Abc_ZddPermProduct(p *Abc_ZddMan, a, b int32) int32 {
	if a == 0 || b == 0 {
		return 0
	}
	if a == 1 {
		return b
	}
	if b == 1 {
		return a
	}
	if r := zddRecall(p, ABC_ZDD_OPER_PERM_PROD, a, b); r >= 0 {
		return r
	}
	B := p.pObjs[b]
	no := Abc_ZddPermProduct(p, a, B.False)
	yes := Abc_ZddPermProduct(p, a, B.True)
	yes = Abc_ZddPerm(p, yes, B.Var)
	return zddRemember(p, ABC_ZDD_OPER_PERM_PROD, a, b, Abc_ZddUnion(p, no, yes))
}
func Abc_ZddPermPrint(perm []int32, size int32) {
	fmt.Print("{")
	for _, v := range perm[:size] {
		fmt.Printf(" %2d", v)
	}
	fmt.Println(" }")
}
func Abc_ZddCombPrint(comb []int32, n int32) {
	if n == 0 {
		fmt.Print("Empty set")
	}
	for _, v := range comb[:n] {
		fmt.Printf("(%d %d)", uint32(v)>>16, uint32(v)&65535)
	}
	fmt.Println()
}
func Abc_ZddPerm2Comb(perm []int32, size int32, comb []int32) int32 {
	perm = perm[:size]
	if size > 65536 {
		panic("permutation too large for packed transpositions")
	}
	seen := make([]bool, size)
	for _, v := range perm {
		if v < 0 || v >= size || seen[v] {
			panic("invalid permutation")
		}
		seen[v] = true
	}
	count := int32(0)
	for i := int32(0); i < size; i++ {
		if perm[i] == i {
			continue
		}
		j := i + 1
		for perm[j] != i {
			j++
		}
		comb[count] = int32(uint32(i)<<16 | uint32(j))
		count++
		perm[i], perm[j] = perm[j], perm[i]
	}
	return count
}
func Abc_ZddComb2Perm(comb []int32, n int32, perm []int32, size int32) {
	comb = comb[:n]
	perm = perm[:size]
	for i := range perm {
		perm[i] = int32(i)
	}
	for i := n - 1; i >= 0; i-- {
		v := uint32(comb[i])
		a, b := v>>16, v&65535
		if a >= uint32(size) || b >= uint32(size) {
			panic("invalid transposition")
		}
		perm[a], perm[b] = perm[b], perm[a]
	}
}
func Abc_ZddPermCombTest() {
	perm := []int32{6, 5, 7, 0, 3, 2, 1, 8, 9, 4}
	comb := make([]int32, len(perm))
	Abc_ZddPermPrint(perm, int32(len(perm)))
	n := Abc_ZddPerm2Comb(perm, int32(len(perm)), comb)
	Abc_ZddCombPrint(comb, n)
	Abc_ZddComb2Perm(comb, n, perm, int32(len(perm)))
	Abc_ZddPermPrint(perm, int32(len(perm)))
}
func Abc_ZddPrint_rec(p *Abc_ZddMan, a int32, path []int32, size int32) {
	if a == 0 {
		return
	}
	if a == 1 {
		if p.nPermSize == 0 {
			Abc_ZddPermPrint(path, size)
			return
		}
		comb := make([]int32, size)
		perm := make([]int32, p.nPermSize)
		for i, v := range path[:size] {
			comb[i] = int32(uint32(p.pV2TI[v])<<16 | uint32(p.pV2TJ[v]))
		}
		Abc_ZddCombPrint(comb, size)
		Abc_ZddComb2Perm(comb, size, perm, p.nPermSize)
		Abc_ZddPermPrint(perm, p.nPermSize)
		return
	}
	A := p.pObjs[a]
	Abc_ZddPrint_rec(p, A.False, path, size)
	path[size] = A.Var
	Abc_ZddPrint_rec(p, A.True, path, size+1)
}
func Abc_ZddPrint(p *Abc_ZddMan, a int32) { Abc_ZddPrint_rec(p, a, make([]int32, p.nVars), 0) }
func Abc_ZddPrintTest(p *Abc_ZddMan) {
	sets := [][]int32{{5, 0, 2, 10, 7}, {3, 11, 10, 7, 2}, {0, 2, 5, 10, 7}}
	var union int32
	for _, set := range sets {
		Abc_ZddPermPrint(set, int32(len(set)))
		union = Abc_ZddUnion(p, union, Abc_ZddBuildSet(p, set, int32(len(set))))
	}
	fmt.Println("Resulting set:")
	Abc_ZddPrint(p, union)
	fmt.Printf("Nodes = %d.   Path = %d.\n", Abc_ZddCountNodes(p, union), Abc_ZddCountPaths(p, union))
}

// Abc_ZddGiaGraph is dependency-neutral to avoid an extra/gia import cycle.
type Abc_ZddGiaGraph interface {
	ObjectCount() int
	InputObjectIDs() []int32
	AndFanins(int32) (int32, int32, bool)
}

func zddGiaCuts(graph Abc_ZddGiaGraph) (*Abc_ZddMan, []int32) {
	n := graph.ObjectCount()
	if n < 0 || int64(n) > 1<<30 {
		panic("invalid graph size")
	}
	p := Abc_ZddManAlloc(int32(n), 1<<24)
	values := make([]int32, n)
	if n > 0 {
		values[0] = 1
	}
	for _, id := range graph.InputObjectIDs() {
		values[id] = Abc_ZddIthVar(id)
	}
	var roots []int32
	for id := 0; id < n; id++ {
		a, b, ok := graph.AndFanins(int32(id))
		if !ok {
			continue
		}
		if a < 0 || b < 0 || a >= int32(id) || b >= int32(id) {
			panic("non-topological graph fanin")
		}
		r := Abc_ZddDotMinProduct6(p, values[a], values[b])
		r = Abc_ZddUnion(p, r, Abc_ZddIthVar(int32(id)))
		values[id] = r
		roots = append(roots, r)
	}
	return p, roots
}
func Abc_ZddGiaTest(graph Abc_ZddGiaGraph) {
	p, roots := zddGiaCuts(graph)
	defer Abc_ZddManFree(p)
	var paths uint64
	for _, r := range roots {
		count := Abc_ZddCountPaths(p, r)
		if ^uint64(0)-paths < count {
			panic("ZDD graph path count overflow")
		}
		paths += count
	}
	fmt.Printf("Paths = %d.  Shared nodes = %d.\n", paths, Abc_ZddCountNodesArray(p, roots))
}
func Abc_ZddPermTestInt(p *Abc_ZddMan) {
	permutations := [][]int32{{1, 0, 2, 4, 3}, {1, 2, 4, 0, 3}, {0, 3, 2, 1, 4}}
	comb := make([]int32, 5)
	var union int32
	for _, perm := range permutations {
		Abc_ZddPermPrint(perm, 5)
		n := Abc_ZddPerm2Comb(perm, 5, comb)
		for i := int32(0); i < n; i++ {
			v := uint32(comb[i])
			comb[i] = Abc_ZddVarIJ(p, int32(v>>16), int32(v&65535))
		}
		union = Abc_ZddUnion(p, union, Abc_ZddBuildSet(p, comb, n))
	}
	Abc_ZddPrint(p, union)
	union = Abc_ZddPerm(p, union, Abc_ZddVarIJ(p, 3, 4))
	Abc_ZddPrint(p, union)
}
func Abc_ZddPermTest() {
	p := Abc_ZddManAlloc(10, 1<<20)
	Abc_ZddManCreatePerms(p, 5)
	defer Abc_ZddManFree(p)
	Abc_ZddPermTestInt(p)
}

var zddCubeTurns = [3][9][2]int32{{{3, 5}, {3, 17}, {3, 15}, {1, 6}, {1, 16}, {1, 14}, {2, 4}, {2, 18}, {2, 13}}, {{2, 14}, {2, 24}, {2, 12}, {3, 13}, {3, 23}, {3, 10}, {1, 15}, {1, 22}, {1, 11}}, {{1, 10}, {1, 7}, {1, 4}, {3, 12}, {3, 9}, {3, 6}, {2, 11}, {2, 8}, {2, 5}}}

func zddCubeStates(rounds int32) (*Abc_ZddMan, int32) {
	p := Abc_ZddManAlloc(276, 1<<27)
	Abc_ZddManCreatePerms(p, 24)
	turns := int32(1)
	for _, swaps := range zddCubeTurns {
		perm := make([]int32, 24)
		for i := range perm {
			perm[i] = int32(i)
		}
		for _, pair := range swaps {
			a, b := pair[0]-1, pair[1]-1
			perm[a], perm[b] = perm[b], perm[a]
		}
		comb := make([]int32, 24)
		n := Abc_ZddPerm2Comb(perm, 24, comb)
		for i := int32(0); i < n; i++ {
			v := uint32(comb[i])
			comb[i] = Abc_ZddVarIJ(p, int32(v>>16), int32(v&65535))
		}
		one := Abc_ZddBuildSet(p, comb, n)
		two := Abc_ZddPermProduct(p, one, one)
		three := Abc_ZddPermProduct(p, two, one)
		turns = Abc_ZddUnion(p, turns, Abc_ZddUnion(p, one, Abc_ZddUnion(p, two, three)))
	}
	all := int32(1)
	for round := int32(0); round < rounds; round++ {
		next := Abc_ZddPermProduct(p, all, turns)
		if next == all {
			break
		}
		all = next
	}
	return p, all
}
func Abc_EnumerateCubeStatesZdd() {
	started := time.Now()
	p, all := zddCubeStates(100)
	defer Abc_ZddManFree(p)
	fmt.Printf("Enumerating states of 2x2x2 cube. Paths = %d. Nodes = %d. Time = %v\n", Abc_ZddCountPaths(p, all), Abc_ZddCountNodes(p, all), time.Since(started))
}

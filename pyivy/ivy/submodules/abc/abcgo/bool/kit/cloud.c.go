package kit

import (
	"fmt"
	"sort"
	"strings"
)

func cloudNodePair(dd *CloudManager, v CloudVar, t, e *CloudNode) *CloudNode {
	n := &CloudNode{s: dd.nSignCur, v: v, t: t, e: e, owner: dd, id: uint32(len(dd.tUnique)) * 2}
	c := &CloudNode{s: dd.nSignCur, v: v, t: t, e: e, owner: dd, id: n.id + 1, complement: true, regular: n, opposite: n}
	n.regular = n
	n.opposite = c
	dd.tUnique = append(dd.tUnique, n)
	return n
}
func Cloud_Init(nVars, nBits int32) *CloudManager {
	if nVars < 0 || nVars > 100000 || nBits < 0 || nBits >= 31 {
		panic("cloud: invalid manager size")
	}
	if nBits == 0 {
		nBits = CLOUD_NODE_BITS
	}
	dd := &CloudManager{nVars: nVars, bitsNode: nBits, nNodesLimit: 1 << uint(nBits), cacheLimit: 1 << uint(max(0, nBits-4)), nSignCur: 1, unique: make(map[cloudKey]*CloudNode), andCache: make(map[cloudPair]*CloudNode)}
	dd.one = cloudNodePair(dd, CLOUD_CONST_INDEX, nil, nil)
	dd.zero = Cloud_Not(dd.one)
	dd.vars = make([]*CloudNode, nVars)
	for i := int32(0); i < nVars; i++ {
		dd.vars[i] = Cloud_MakeNode(dd, uint32(i), dd.one, dd.zero)
		if dd.vars[i] == nil {
			panic("cloud: node limit too small for variables")
		}
	}
	return dd
}
func Cloud_Quit(dd *CloudManager) {
	if dd == nil {
		return
	}
	dd.vars = nil
	dd.ppNodes = nil
	dd.tUnique = nil
	dd.unique = nil
	dd.andCache = nil
	dd.one = nil
	dd.zero = nil
}
func Cloud_Restart(dd *CloudManager) {
	next := Cloud_Init(dd.nVars, dd.bitsNode)
	sign := dd.nSignCur + 1
	*dd = *next
	dd.nSignCur = sign
	for _, n := range dd.tUnique {
		n.owner = dd
		n.opposite.owner = dd
		n.s = sign
		n.opposite.s = sign
	}
}
func Cloud_CacheAllocate(dd *CloudManager, oper CloudOper, logratio int32) {
	if oper < 0 || oper >= 4 || logratio < 0 {
		panic("cloud: invalid cache configuration")
	}
	// All Boolean operations use the AND backend; only it needs a computed
	// cache. Apply its requested memory ratio to the actual map budget.
	if oper == CLOUD_OPER_AND {
		dd.cacheLimit = 1 << uint(max(0, dd.bitsNode-logratio))
		clear(dd.andCache)
	}
}
func cloudCacheAllocate(dd *CloudManager, oper CloudOper) { Cloud_CacheAllocate(dd, oper, 4) }
func cloudCheck(dd *CloudManager, n *CloudNode) {
	if n == nil || n.owner != dd || n.s != dd.nSignCur {
		panic("cloud: foreign or expired BDD node")
	}
}
func Cloud_MakeNode(dd *CloudManager, v CloudVar, t, e *CloudNode) *CloudNode {
	cloudCheck(dd, t)
	cloudCheck(dd, e)
	if t == e {
		return t
	}
	if v >= uint32(dd.nVars) || v >= Cloud_V(t) || v >= Cloud_V(e) {
		panic("cloud: unordered BDD node")
	}
	complement := Cloud_IsComplement(t)
	if complement {
		t = Cloud_Not(t)
		e = Cloud_Not(e)
	}
	n := cloudMakeNode(dd, v, t, e)
	return Cloud_NotCond(n, complement)
}
func cloudMakeNode(dd *CloudManager, v CloudVar, t, e *CloudNode) *CloudNode {
	key := cloudKey{v, t, e}
	if n := dd.unique[key]; n != nil {
		dd.nUniqueHits++
		return n
	}
	dd.nUniqueMisses++
	if int32(len(dd.tUnique))+1 >= dd.nNodesLimit {
		return nil
	}
	n := cloudNodePair(dd, v, t, e)
	dd.unique[key] = n
	return n
}
func cloudCofactor(n *CloudNode, v uint32) (*CloudNode, *CloudNode) {
	if Cloud_V(n) != v {
		return n, n
	}
	return Cloud_NotCond(Cloud_E(n), n.complement), Cloud_NotCond(Cloud_T(n), n.complement)
}
func cloudBddAnd(dd *CloudManager, f, g *CloudNode) *CloudNode {
	if f == dd.zero || g == dd.zero || f == Cloud_Not(g) {
		return dd.zero
	}
	if f == dd.one {
		return g
	}
	if g == dd.one || f == g {
		return f
	}
	if f.id > g.id {
		f, g = g, f
	}
	key := cloudPair{f, g}
	if r := dd.andCache[key]; r != nil {
		dd.nCacheHits++
		return r
	}
	dd.nCacheMisses++
	v := min(Cloud_V(f), Cloud_V(g))
	f0, f1 := cloudCofactor(f, v)
	g0, g1 := cloudCofactor(g, v)
	low := cloudBddAnd(dd, f0, g0)
	if low == nil {
		return nil
	}
	high := cloudBddAnd(dd, f1, g1)
	if high == nil {
		return nil
	}
	r := Cloud_MakeNode(dd, v, high, low)
	if r != nil {
		if len(dd.andCache) >= dd.cacheLimit {
			clear(dd.andCache)
		}
		dd.andCache[key] = r
	}
	return r
}
func cloudBddAnd_gate(dd *CloudManager, f, g *CloudNode) *CloudNode { return Cloud_bddAnd(dd, f, g) }
func Cloud_bddAnd(dd *CloudManager, f, g *CloudNode) *CloudNode {
	cloudCheck(dd, f)
	cloudCheck(dd, g)
	return cloudBddAnd(dd, f, g)
}
func Cloud_bddOr(dd *CloudManager, f, g *CloudNode) *CloudNode {
	return Cloud_Not(Cloud_bddAnd(dd, Cloud_Not(f), Cloud_Not(g)))
}
func Cloud_bddXor(dd *CloudManager, f, g *CloudNode) *CloudNode {
	a := Cloud_bddAnd(dd, f, Cloud_Not(g))
	if a == nil {
		return nil
	}
	b := Cloud_bddAnd(dd, Cloud_Not(f), g)
	if b == nil {
		return nil
	}
	return Cloud_bddOr(dd, a, b)
}
func cloudWalk(n *CloudNode, seen map[*CloudNode]bool, visit func(*CloudNode)) {
	n = Cloud_Regular(n)
	if seen[n] {
		return
	}
	seen[n] = true
	if !Cloud_IsConstant(n) {
		cloudWalk(n.t, seen, visit)
		cloudWalk(n.e, seen, visit)
	}
	visit(n)
}
func cloudClearMark(dd *CloudManager, n *CloudNode) {
	cloudWalk(n, map[*CloudNode]bool{}, func(x *CloudNode) { cloudNodeUnmark(x) })
}
func cloudSupport(dd *CloudManager, n *CloudNode, support []int32) {
	cloudWalk(n, map[*CloudNode]bool{}, func(x *CloudNode) {
		if !Cloud_IsConstant(x) {
			support[Cloud_V(x)] = 1
		}
	})
}
func Cloud_Support(dd *CloudManager, n *CloudNode) *CloudNode {
	cloudCheck(dd, n)
	support := make([]int32, dd.nVars)
	cloudSupport(dd, n, support)
	r := dd.one
	for i := dd.nVars - 1; i >= 0; i-- {
		if support[i] != 0 {
			r = Cloud_bddAnd(dd, r, dd.vars[i])
			if r == nil {
				return nil
			}
		}
	}
	return r
}
func Cloud_SupportSize(dd *CloudManager, n *CloudNode) int32 {
	cloudCheck(dd, n)
	support := make([]int32, dd.nVars)
	cloudSupport(dd, n, support)
	var count int32
	for _, x := range support {
		count += x
	}
	return count
}
func cloudDagSize(dd *CloudManager, n *CloudNode) int32 { return Cloud_DagSize(dd, n) }
func Cloud_DagSize(dd *CloudManager, n *CloudNode) int32 {
	cloudCheck(dd, n)
	seen := map[*CloudNode]bool{}
	cloudWalk(n, seen, func(*CloudNode) {})
	return int32(len(seen))
}
func Cloud_DagCollect_rec(dd *CloudManager, n *CloudNode, counter *int32) int32 {
	before := *counter
	seen := map[*CloudNode]bool{}
	cloudWalk(n, seen, func(x *CloudNode) { dd.ppNodes = append(dd.ppNodes, x); *counter++ })
	return *counter - before
}
func Cloud_DagCollect(dd *CloudManager, n *CloudNode) int32 {
	cloudCheck(dd, n)
	dd.ppNodes = dd.ppNodes[:0]
	var count int32
	return Cloud_DagCollect_rec(dd, n, &count)
}
func Cloud_SharingSize(dd *CloudManager, nodes []*CloudNode, count int32) int32 {
	seen := map[*CloudNode]bool{}
	for _, n := range nodes[:count] {
		cloudCheck(dd, n)
		cloudWalk(n, seen, func(*CloudNode) {})
	}
	return int32(len(seen))
}
func Cloud_GetOneCube(dd *CloudManager, n *CloudNode) *CloudNode {
	cloudCheck(dd, n)
	if Cloud_IsConstant(n) {
		return n
	}
	v := Cloud_V(n)
	low, high := cloudCofactor(n, v)
	r := Cloud_GetOneCube(dd, low)
	if r == nil {
		return nil
	}
	if r != dd.zero {
		return Cloud_bddAnd(dd, r, Cloud_Not(dd.vars[v]))
	}
	r = Cloud_GetOneCube(dd, high)
	if r == nil {
		return nil
	}
	return Cloud_bddAnd(dd, r, dd.vars[v])
}
func Cloud_Evaluate(dd *CloudManager, n *CloudNode, values []bool) bool {
	cloudCheck(dd, n)
	if len(values) < int(dd.nVars) {
		panic("cloud: short assignment")
	}
	for !Cloud_IsConstant(n) {
		low, high := cloudCofactor(n, Cloud_V(n))
		if values[Cloud_V(n)] {
			n = high
		} else {
			n = low
		}
	}
	return n == dd.one
}
func Cloud_bddPrint(dd *CloudManager, n *CloudNode) {
	path := make([]byte, dd.nVars)
	for i := range path {
		path[i] = '-'
	}
	var walk func(*CloudNode)
	walk = func(x *CloudNode) {
		if x == dd.zero {
			return
		}
		if x == dd.one {
			fmt.Printf("%s 1\n", path)
			return
		}
		v := Cloud_V(x)
		low, high := cloudCofactor(x, v)
		path[v] = '0'
		walk(low)
		path[v] = '1'
		walk(high)
		path[v] = '-'
	}
	cloudCheck(dd, n)
	walk(n)
}
func Cloud_bddPrintCube(dd *CloudManager, n *CloudNode) {
	cloudCheck(dd, n)
	var literals []string
	for !Cloud_IsConstant(n) {
		v := Cloud_V(n)
		low, high := cloudCofactor(n, v)
		if low == dd.zero {
			literals = append(literals, fmt.Sprintf("x%d", v))
			n = high
		} else if high == dd.zero {
			literals = append(literals, fmt.Sprintf("!x%d", v))
			n = low
		} else {
			panic("cloud: function is not a cube")
		}
	}
	if n == dd.zero {
		fmt.Println("0")
		return
	}
	fmt.Println(strings.Join(literals, " "))
}
func Cloud_PrintInfo(dd *CloudManager) {
	fmt.Printf("Cloud: vars=%d nodes=%d limit=%d unique-hits=%d unique-misses=%d cache-hits=%d cache-misses=%d\n", dd.nVars, len(dd.tUnique), dd.nNodesLimit, dd.nUniqueHits, dd.nUniqueMisses, dd.nCacheHits, dd.nCacheMisses)
}
func Cloud_PrintHashTable(dd *CloudManager) {
	ids := make([]int, 0, len(dd.tUnique))
	byID := map[int]*CloudNode{}
	for _, n := range dd.tUnique {
		ids = append(ids, int(n.id))
		byID[int(n.id)] = n
	}
	sort.Ints(ids)
	for _, id := range ids {
		n := byID[id]
		if Cloud_IsConstant(n) {
			fmt.Printf("%d: 1\n", id)
		} else {
			fmt.Printf("%d: x%d ? %d : %d\n", id, n.v, n.t.id, n.e.id)
		}
	}
}

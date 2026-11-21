package kit

// Decomposition graph edges carry a node index and an explicit complement bit.
// Go nodes are individually allocated so their addresses stay stable on append.
type Kit_Edge_t struct{ fCompl, Node uint32 }
type Kit_Node_t struct {
	pFunc          any
	eEdge0, eEdge1 Kit_Edge_t
	index, iFunc   int32
	Level          uint32
	printing       uint8 // Original OR and input complement bits, in that order.
	latches        [3]uint8
}
type Kit_Graph_t struct {
	fConst, nLeaves int32
	pNodes          []*Kit_Node_t
	eRoot           Kit_Edge_t
}

func Kit_EdgeCreate(node, compl int32) Kit_Edge_t {
	return Kit_Edge_t{uint32(compl) & 1, uint32(node)}.normalized()
}
func (e Kit_Edge_t) normalized() Kit_Edge_t {
	e.Node &= 1<<30 - 1
	return e
}
func Kit_EdgeToInt(e Kit_Edge_t) uint32  { return e.normalized().Node<<1 | e.fCompl }
func Kit_IntToEdge(e uint32) Kit_Edge_t  { return Kit_EdgeCreate(int32(e>>1), int32(e&1)) }
func Kit_EdgeToInt_(e Kit_Edge_t) uint32 { return Kit_EdgeToInt(e) }
func Kit_IntToEdge_(e uint32) Kit_Edge_t { return Kit_IntToEdge(e) }
func kitFlag(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func Kit_GraphIsConst(g *Kit_Graph_t) int32             { return g.fConst }
func Kit_GraphIsConst0(g *Kit_Graph_t) int32            { return kitFlag(g.fConst != 0 && g.eRoot.fCompl != 0) }
func Kit_GraphIsConst1(g *Kit_Graph_t) int32            { return kitFlag(g.fConst != 0 && g.eRoot.fCompl == 0) }
func Kit_GraphIsComplement(g *Kit_Graph_t) int32        { return int32(g.eRoot.fCompl) }
func Kit_GraphIsVar(g *Kit_Graph_t) int32               { return kitFlag(g.eRoot.Node < uint32(g.nLeaves)) }
func Kit_GraphComplement(g *Kit_Graph_t)                { g.eRoot.fCompl ^= 1 }
func Kit_GraphSetRoot(g *Kit_Graph_t, e Kit_Edge_t)     { g.eRoot = e.normalized() }
func Kit_GraphLeaveNum(g *Kit_Graph_t) int32            { return g.nLeaves }
func Kit_GraphNodeNum(g *Kit_Graph_t) int32             { return int32(len(g.pNodes)) - g.nLeaves }
func Kit_GraphNode(g *Kit_Graph_t, i int32) *Kit_Node_t { return g.pNodes[i] }
func Kit_GraphNodeLast(g *Kit_Graph_t) *Kit_Node_t      { return g.pNodes[int32(len(g.pNodes))-1] }
func Kit_GraphNodeInt(g *Kit_Graph_t, n *Kit_Node_t) int32 {
	if id := kitNodeIndex(g, n); id >= 0 {
		return id
	}
	panic("node does not belong to graph")
}
func kitNodeIndex(g *Kit_Graph_t, n *Kit_Node_t) int32 {
	if n != nil && n.index >= 0 && int(n.index) < len(g.pNodes) && g.pNodes[n.index] == n {
		return n.index
	}
	// A caller may overwrite an exported node by value, including its index.
	for i, p := range g.pNodes {
		if p == n {
			return int32(i)
		}
	}
	return -1
}
func Kit_GraphNodeIsVar(g *Kit_Graph_t, n *Kit_Node_t) int32 {
	return kitFlag(Kit_GraphNodeInt(g, n) < g.nLeaves)
}
func Kit_GraphVar(g *Kit_Graph_t) *Kit_Node_t {
	if Kit_GraphIsVar(g) == 0 {
		panic("graph is not a variable")
	}
	return g.pNodes[g.eRoot.Node]
}
func Kit_GraphVarInt(g *Kit_Graph_t) int32 { Kit_GraphVar(g); return int32(g.eRoot.Node) }
func Kit_GraphNodeFanin0(g *Kit_Graph_t, n *Kit_Node_t) *Kit_Node_t {
	if Kit_GraphNodeIsVar(g, n) != 0 {
		return nil
	}
	return g.pNodes[n.eEdge0.Node]
}
func Kit_GraphNodeFanin1(g *Kit_Graph_t, n *Kit_Node_t) *Kit_Node_t {
	if Kit_GraphNodeIsVar(g, n) != 0 {
		return nil
	}
	return g.pNodes[n.eEdge1.Node]
}
func Kit_GraphRootLevel(g *Kit_Graph_t) int32 {
	if g.fConst != 0 {
		return 0
	}
	return int32(g.pNodes[g.eRoot.Node].Level)
}
func Kit_GraphCreate(n int32) *Kit_Graph_t {
	if n < 0 {
		panic("negative leaf count")
	}
	g := &Kit_Graph_t{nLeaves: n, pNodes: make([]*Kit_Node_t, n, 2*n+50)}
	for i := range g.pNodes {
		g.pNodes[i] = &Kit_Node_t{index: int32(i)}
	}
	return g
}
func Kit_GraphCreateConst0() *Kit_Graph_t {
	return &Kit_Graph_t{fConst: 1, eRoot: Kit_Edge_t{fCompl: 1}}
}
func Kit_GraphCreateConst1() *Kit_Graph_t { return &Kit_Graph_t{fConst: 1} }
func Kit_GraphCreateLeaf(i, n, c int32) *Kit_Graph_t {
	if i < 0 || i >= n {
		panic("invalid graph leaf")
	}
	g := Kit_GraphCreate(n)
	g.eRoot = Kit_EdgeCreate(i, c)
	return g
}
func Kit_GraphFree(g *Kit_Graph_t) {
	if g != nil {
		g.pNodes = nil
		g.nLeaves = 0
	}
}
func Kit_GraphAppendNode(g *Kit_Graph_t) *Kit_Node_t {
	n := &Kit_Node_t{index: int32(len(g.pNodes))}
	g.pNodes = append(g.pNodes, n)
	return n
}
func Kit_GraphAddNodeAnd(g *Kit_Graph_t, a, b Kit_Edge_t) Kit_Edge_t {
	a, b = a.normalized(), b.normalized()
	if int(a.Node) >= len(g.pNodes) || int(b.Node) >= len(g.pNodes) {
		panic("fanin outside graph")
	}
	level := max(g.pNodes[a.Node].Level, g.pNodes[b.Node].Level) + 1
	n := Kit_GraphAppendNode(g)
	n.eEdge0 = a
	n.eEdge1 = b
	n.printing = uint8(a.fCompl<<1 | b.fCompl<<2)
	n.Level = level
	return Kit_EdgeCreate(int32(len(g.pNodes))-1, 0)
}
func Kit_GraphAddNodeOr(g *Kit_Graph_t, a, b Kit_Edge_t) Kit_Edge_t {
	originalA, originalB := a.fCompl, b.fCompl
	a.fCompl ^= 1
	b.fCompl ^= 1
	e := Kit_GraphAddNodeAnd(g, a, b)
	n := g.pNodes[e.Node]
	n.printing = uint8(1 | originalA<<1 | originalB<<2)
	e.fCompl = 1
	return e
}
func Kit_GraphAddNodeXor(g *Kit_Graph_t, a, b Kit_Edge_t, kind int32) Kit_Edge_t {
	var x, y Kit_Edge_t
	if kind == 0 {
		a.fCompl ^= 1
		x = Kit_GraphAddNodeAnd(g, a, b)
		a.fCompl ^= 1
		b.fCompl ^= 1
		y = Kit_GraphAddNodeAnd(g, a, b)
	} else {
		x = Kit_GraphAddNodeAnd(g, a, b)
		a.fCompl ^= 1
		b.fCompl ^= 1
		y = Kit_GraphAddNodeAnd(g, a, b)
	}
	e := Kit_GraphAddNodeOr(g, x, y)
	if kind != 0 {
		e.fCompl ^= 1
	}
	return e
}
func Kit_GraphAddNodeMux(g *Kit_Graph_t, c, t, e Kit_Edge_t, kind int32) Kit_Edge_t {
	if kind != 0 {
		t.fCompl ^= 1
		e.fCompl ^= 1
	}
	a := Kit_GraphAddNodeAnd(g, c, t)
	c.fCompl ^= 1
	b := Kit_GraphAddNodeAnd(g, c, e)
	result := Kit_GraphAddNodeOr(g, a, b)
	if kind != 0 {
		result.fCompl ^= 1
	}
	return result
}
func Kit_GraphToTruth(g *Kit_Graph_t) uint32 {
	if g.nLeaves > 5 {
		panic("single-word graph truth requires at most five variables")
	}
	if g.fConst != 0 {
		if g.eRoot.fCompl != 0 {
			return 0
		}
		return ^uint32(0)
	}
	vars := [5]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	if Kit_GraphIsVar(g) != 0 {
		word := vars[g.eRoot.Node]
		if g.eRoot.fCompl != 0 {
			return ^word
		}
		return word
	}
	for i := int32(0); i < g.nLeaves; i++ {
		g.pNodes[i].iFunc = int32(vars[i])
	}
	for i := g.nLeaves; i < int32(len(g.pNodes)); i++ {
		n := g.pNodes[i]
		a, b := uint32(g.pNodes[n.eEdge0.Node].iFunc), uint32(g.pNodes[n.eEdge1.Node].iFunc)
		if n.eEdge0.fCompl != 0 {
			a = ^a
		}
		if n.eEdge1.fCompl != 0 {
			b = ^b
		}
		n.iFunc = int32(a & b)
	}
	result := uint32(g.pNodes[g.eRoot.Node].iFunc)
	if g.eRoot.fCompl != 0 {
		result = ^result
	}
	return result
}

// Distances are stored as length plus one, so zero denotes an unreachable node.
// Every reachable node in a graph built with Kit_GraphAddNodeAnd follows both
// fanins. Descending order therefore settles the longest path before use.
func kitPropagateDepths(g *Kit_Graph_t, root int32, distances []int32) bool {
	distances[root] = 1
	for i := root; i >= g.nLeaves; i-- {
		d := distances[i]
		if d == 0 {
			continue
		}
		n := g.pNodes[i]
		a, b := n.eEdge0.Node, n.eEdge1.Node
		if a >= uint32(i) || b >= uint32(i) {
			return false
		}
		distances[a] = max(distances[a], d+1)
		distances[b] = max(distances[b], d+1)
	}
	return true
}

// Node assignment can introduce acyclic forward references. Memoize the
// original fanin traversal for that case, visiting only the requested cone.
func kitDepthByFanins(g *Kit_Graph_t, id uint32, target *Kit_Node_t, level bool, memo []int32) int32 {
	n := g.pNodes[id]
	if n == target || (level && id < uint32(g.nLeaves)) {
		return 0
	}
	if id < uint32(g.nLeaves) {
		return -100
	}
	if saved := memo[id]; saved != 0 {
		return saved - 1
	}
	d := max(kitDepthByFanins(g, n.eEdge0.Node, target, level, memo), kitDepthByFanins(g, n.eEdge1.Node, target, level, memo))
	if d != -100 {
		d++
	}
	memo[id] = d + 1
	return d
}

// Try the reachable cone before allocating a table for a high-numbered root.
// Abandon this walk after 31 entries so reconvergence cannot multiply the work.
func kitSmallConeDepth(g *Kit_Graph_t, root int32, target *Kit_Node_t, level bool, leaves []int32) (int32, bool) {
	type entry struct {
		id    uint32
		depth int32
	}
	var queue [31]entry
	queue[0].id = uint32(root)
	end, best := 1, int32(-100)
	for i := 0; i < end; i++ {
		id, d := queue[i].id, queue[i].depth
		node := g.pNodes[id]
		if node == target {
			best = max(best, d)
			continue
		}
		if id < uint32(g.nLeaves) {
			if level {
				best = max(best, d)
			}
			if int(id) < len(leaves) {
				leaves[id] = max(leaves[id], d)
			}
			continue
		}
		if end+2 > len(queue) {
			return 0, false
		}
		queue[end] = entry{node.eEdge0.Node, d + 1}
		queue[end+1] = entry{node.eEdge1.Node, d + 1}
		end += 2
	}
	return best, true
}

func kitPathDepth(g *Kit_Graph_t, root, target *Kit_Node_t, level bool) int32 {
	rootID := Kit_GraphNodeInt(g, root)
	if root == target || (level && rootID < g.nLeaves) {
		return 0
	}
	if rootID < g.nLeaves {
		return -100
	}
	// A shallow cone can follow a large unrelated graph.
	if root.eEdge0.Node < uint32(g.nLeaves) && root.eEdge1.Node < uint32(g.nLeaves) {
		if level || g.pNodes[root.eEdge0.Node] == target || g.pNodes[root.eEdge1.Node] == target {
			return 1
		}
		return -100
	}
	var local [80]int32
	distances := local[:]
	if int(rootID) >= len(distances) {
		if d, done := kitSmallConeDepth(g, rootID, target, level, nil); done {
			return d
		}
		distances = make([]int32, int(rootID)+1)
	}
	if !kitPropagateDepths(g, rootID, distances) {
		return kitDepthByFanins(g, uint32(rootID), target, level, make([]int32, len(g.pNodes)))
	}
	if level {
		longest := int32(0)
		for _, d := range distances[:g.nLeaves] {
			longest = max(d, longest)
		}
		return longest - 1
	}
	if target != nil {
		if id := kitNodeIndex(g, target); id >= 0 && id <= rootID && distances[id] != 0 {
			return distances[id] - 1
		}
	}
	return -100
}

func Kit_GraphLeafDepth_rec(g *Kit_Graph_t, n, leaf *Kit_Node_t) int32 {
	if n == leaf {
		return 0
	}
	return kitPathDepth(g, n, leaf, false)
}
func Kit_GraphLevelNum_rec(g *Kit_Graph_t, n *Kit_Node_t) int32 {
	return kitPathDepth(g, n, nil, true)
}

// Kit_GraphLeafDepths fills dst with the longest distance from each leaf to
// root. An unused leaf has depth -100, matching Kit_GraphLeafDepth_rec.
// A constant graph has a nil root and all requested leaves are unused.
func Kit_GraphLeafDepths(g *Kit_Graph_t, root *Kit_Node_t, dst []int32) {
	if len(dst) == 0 {
		return
	}
	if g.fConst != 0 {
		for i := range dst {
			dst[i] = -100
		}
		return
	}
	id := Kit_GraphNodeInt(g, root)
	var local [80]int32
	distances := local[:]
	if int(id) >= len(distances) {
		for i := range dst {
			dst[i] = -100
		}
		if _, done := kitSmallConeDepth(g, id, nil, false, dst); done {
			return
		}
		distances = make([]int32, int(id)+1)
	}
	ordered := kitPropagateDepths(g, id, distances)
	for i := range dst {
		dst[i] = -100
		if i >= int(g.nLeaves) {
			continue
		}
		if !ordered {
			dst[i] = kitDepthByFanins(g, uint32(id), g.pNodes[i], false, make([]int32, len(g.pNodes)))
		} else if i <= int(id) && distances[i] != 0 {
			dst[i] = distances[i] - 1
		}
	}
}

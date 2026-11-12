package kit

// Decomposition graph edges carry a node index and an explicit complement bit.
// Go nodes are individually allocated so their addresses stay stable on append.
type Kit_Edge_t struct{ fCompl, Node uint32 }
type Kit_Node_t struct {
	eEdge0, eEdge1            Kit_Edge_t
	iFunc                     int32
	pFunc                     any
	Level                     uint32
	fNodeOr, fCompl0, fCompl1 uint32
	nLat0, nLat1, nLat2       uint32
}
type Kit_Graph_t struct {
	fConst, nLeaves int32
	pNodes          []*Kit_Node_t
	eRoot           Kit_Edge_t
}

func Kit_EdgeCreate(node, compl int32) Kit_Edge_t { return Kit_Edge_t{uint32(compl) & 1, uint32(node)} }
func Kit_EdgeToInt(e Kit_Edge_t) uint32           { return e.Node<<1 | e.fCompl }
func Kit_IntToEdge(e uint32) Kit_Edge_t           { return Kit_EdgeCreate(int32(e>>1), int32(e&1)) }
func Kit_EdgeToInt_(e Kit_Edge_t) uint32          { return Kit_EdgeToInt(e) }
func Kit_IntToEdge_(e uint32) Kit_Edge_t          { return Kit_IntToEdge(e) }
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
func Kit_GraphSetRoot(g *Kit_Graph_t, e Kit_Edge_t)     { g.eRoot = e }
func Kit_GraphLeaveNum(g *Kit_Graph_t) int32            { return g.nLeaves }
func Kit_GraphNodeNum(g *Kit_Graph_t) int32             { return int32(len(g.pNodes)) - g.nLeaves }
func Kit_GraphNode(g *Kit_Graph_t, i int32) *Kit_Node_t { return g.pNodes[i] }
func Kit_GraphNodeLast(g *Kit_Graph_t) *Kit_Node_t      { return g.pNodes[int32(len(g.pNodes))-1] }
func Kit_GraphNodeInt(g *Kit_Graph_t, n *Kit_Node_t) int32 {
	for i, p := range g.pNodes {
		if p == n {
			return int32(i)
		}
	}
	panic("node does not belong to graph")
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
		g.pNodes[i] = &Kit_Node_t{}
	}
	return g
}
func Kit_GraphCreateConst0() *Kit_Graph_t { g := Kit_GraphCreateConst1(); g.eRoot.fCompl = 1; return g }
func Kit_GraphCreateConst1() *Kit_Graph_t { g := Kit_GraphCreate(0); g.fConst = 1; return g }
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
	n := &Kit_Node_t{}
	g.pNodes = append(g.pNodes, n)
	return n
}
func Kit_GraphAddNodeAnd(g *Kit_Graph_t, a, b Kit_Edge_t) Kit_Edge_t {
	if int(a.Node) >= len(g.pNodes) || int(b.Node) >= len(g.pNodes) {
		panic("fanin outside graph")
	}
	level := max(g.pNodes[a.Node].Level, g.pNodes[b.Node].Level) + 1
	n := Kit_GraphAppendNode(g)
	n.eEdge0 = a
	n.eEdge1 = b
	n.fCompl0 = a.fCompl
	n.fCompl1 = b.fCompl
	n.Level = level
	return Kit_EdgeCreate(int32(len(g.pNodes))-1, 0)
}
func Kit_GraphAddNodeOr(g *Kit_Graph_t, a, b Kit_Edge_t) Kit_Edge_t {
	originalA, originalB := a.fCompl, b.fCompl
	a.fCompl ^= 1
	b.fCompl ^= 1
	e := Kit_GraphAddNodeAnd(g, a, b)
	n := g.pNodes[e.Node]
	n.fNodeOr = 1
	n.fCompl0 = originalA
	n.fCompl1 = originalB
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
	values := make([]uint32, int32(len(g.pNodes)))
	for i := int32(0); i < g.nLeaves; i++ {
		values[i] = vars[i]
		g.pNodes[i].iFunc = int32(vars[i])
	}
	for i := g.nLeaves; i < int32(len(g.pNodes)); i++ {
		n := g.pNodes[i]
		a, b := values[n.eEdge0.Node], values[n.eEdge1.Node]
		if n.eEdge0.fCompl != 0 {
			a = ^a
		}
		if n.eEdge1.fCompl != 0 {
			b = ^b
		}
		values[i] = a & b
		n.iFunc = int32(values[i])
	}
	result := values[g.eRoot.Node]
	if g.eRoot.fCompl != 0 {
		result = ^result
	}
	return result
}
func Kit_GraphLeafDepth_rec(g *Kit_Graph_t, n, leaf *Kit_Node_t) int32 {
	if n == leaf {
		return 0
	}
	if Kit_GraphNodeIsVar(g, n) != 0 {
		return -100
	}
	d := max(Kit_GraphLeafDepth_rec(g, Kit_GraphNodeFanin0(g, n), leaf), Kit_GraphLeafDepth_rec(g, Kit_GraphNodeFanin1(g, n), leaf))
	if d == -100 {
		return d
	}
	return d + 1
}
func Kit_GraphLevelNum_rec(g *Kit_Graph_t, n *Kit_Node_t) int32 {
	if Kit_GraphNodeIsVar(g, n) != 0 {
		return 0
	}
	return 1 + max(Kit_GraphLevelNum_rec(g, Kit_GraphNodeFanin0(g, n)), Kit_GraphLevelNum_rec(g, Kit_GraphNodeFanin1(g, n)))
}

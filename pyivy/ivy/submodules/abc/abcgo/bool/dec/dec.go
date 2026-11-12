package dec

// C-compatible entry points for decomposition graphs. Node pointers remain
// stable when the graph grows: Go owns individually allocated node objects.
type Dec_Edge_t struct {
	fCompl uint32
	Node   uint32
}
type Dec_Node_t struct {
	eEdge0, eEdge1            Dec_Edge_t
	iFunc                     int32
	pFunc                     any
	Level                     uint32
	fNodeOr, fCompl0, fCompl1 uint32
	nLat0, nLat1, nLat2       uint32
}
type Dec_Graph_t struct {
	fConst               int32
	nLeaves, nSize, nCap int32
	pNodes               []*Dec_Node_t
	eRoot                Dec_Edge_t
}
type Dec_Man_t struct {
	pMvcMem               any
	vCubes, vLits         []int32
	puCanons              []uint16
	pPhases, pPerms, pMap []byte
}

func decBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func Dec_EdgeCreate(node, compl int32) Dec_Edge_t {
	if node < 0 || node >= 1<<30 {
		panic("dec: edge index")
	}
	return Dec_Edge_t{uint32(compl) & 1, uint32(node)}
}
func Dec_EdgeToInt(e Dec_Edge_t) uint32 { return e.Node<<1 | e.fCompl }
func Dec_IntToEdge(e uint32) Dec_Edge_t { return Dec_EdgeCreate(int32(e>>1), int32(e&1)) }

// The historical union variant uses native bitfield layout: complement first.
func Dec_EdgeToInt_(e Dec_Edge_t) uint32 { return Dec_EdgeToInt(e) }
func Dec_IntToEdge_(e uint32) Dec_Edge_t { return Dec_IntToEdge(e) }
func Dec_GraphCreate(leaves int32) *Dec_Graph_t {
	if leaves < 0 {
		panic("dec: negative leaves")
	}
	g := &Dec_Graph_t{nLeaves: leaves, nSize: leaves, nCap: 2*leaves + 50, pNodes: make([]*Dec_Node_t, leaves, 2*leaves+50)}
	for i := range g.pNodes {
		g.pNodes[i] = new(Dec_Node_t)
	}
	return g
}
func Dec_GraphCreateConst0() *Dec_Graph_t {
	return &Dec_Graph_t{fConst: 1, eRoot: Dec_Edge_t{fCompl: 1}}
}
func Dec_GraphCreateConst1() *Dec_Graph_t { return &Dec_Graph_t{fConst: 1} }
func Dec_GraphCreateLeaf(i, n, c int32) *Dec_Graph_t {
	g := Dec_GraphCreate(n)
	if i < 0 || i >= n {
		panic("dec: leaf index")
	}
	g.eRoot = Dec_EdgeCreate(i, c)
	return g
}
func Dec_GraphFree(g *Dec_Graph_t) {
	if g != nil {
		g.pNodes = nil
		g.nSize = 0
		g.nCap = 0
	}
}
func Dec_GraphIsConst(g *Dec_Graph_t) int32             { return g.fConst }
func Dec_GraphIsConst0(g *Dec_Graph_t) int32            { return decBool(g.fConst != 0 && g.eRoot.fCompl != 0) }
func Dec_GraphIsConst1(g *Dec_Graph_t) int32            { return decBool(g.fConst != 0 && g.eRoot.fCompl == 0) }
func Dec_GraphIsComplement(g *Dec_Graph_t) int32        { return int32(g.eRoot.fCompl) }
func Dec_GraphComplement(g *Dec_Graph_t)                { g.eRoot.fCompl ^= 1 }
func Dec_GraphLeaveNum(g *Dec_Graph_t) int32            { return g.nLeaves }
func Dec_GraphNodeNum(g *Dec_Graph_t) int32             { return g.nSize - g.nLeaves }
func Dec_GraphNode(g *Dec_Graph_t, i int32) *Dec_Node_t { return g.pNodes[i] }
func Dec_GraphNodeLast(g *Dec_Graph_t) *Dec_Node_t      { return g.pNodes[g.nSize-1] }
func Dec_GraphNodeInt(g *Dec_Graph_t, n *Dec_Node_t) int32 {
	for i, p := range g.pNodes {
		if p == n {
			return int32(i)
		}
	}
	panic("dec: foreign node")
}
func Dec_GraphIsVar(g *Dec_Graph_t) int32 { return decBool(g.eRoot.Node < uint32(g.nLeaves)) }
func Dec_GraphNodeIsVar(g *Dec_Graph_t, n *Dec_Node_t) int32 {
	return decBool(Dec_GraphNodeInt(g, n) < g.nLeaves)
}
func Dec_GraphVar(g *Dec_Graph_t) *Dec_Node_t       { return Dec_GraphNode(g, int32(g.eRoot.Node)) }
func Dec_GraphVarInt(g *Dec_Graph_t) int32          { return int32(g.eRoot.Node) }
func Dec_GraphSetRoot(g *Dec_Graph_t, e Dec_Edge_t) { g.eRoot = e }
func Dec_GraphAppendNode(g *Dec_Graph_t) *Dec_Node_t {
	n := new(Dec_Node_t)
	g.pNodes = append(g.pNodes, n)
	g.nSize = int32(len(g.pNodes))
	g.nCap = int32(cap(g.pNodes))
	return n
}
func Dec_GraphAddNodeAnd(g *Dec_Graph_t, a, b Dec_Edge_t) Dec_Edge_t {
	n := Dec_GraphAppendNode(g)
	n.eEdge0 = a
	n.eEdge1 = b
	n.fCompl0 = a.fCompl
	n.fCompl1 = b.fCompl
	return Dec_EdgeCreate(g.nSize-1, 0)
}
func Dec_GraphAddNodeOr(g *Dec_Graph_t, a, b Dec_Edge_t) Dec_Edge_t {
	e := Dec_GraphAddNodeAnd(g, a, b)
	n := Dec_GraphNodeLast(g)
	n.fNodeOr = 1
	n.eEdge0.fCompl ^= 1
	n.eEdge1.fCompl ^= 1
	e.fCompl = 1
	return e
}
func Dec_GraphAddNodeXor(g *Dec_Graph_t, a, b Dec_Edge_t, kind int32) Dec_Edge_t {
	if kind == 0 {
		na := a
		na.fCompl ^= 1
		nb := b
		nb.fCompl ^= 1
		return Dec_GraphAddNodeOr(g, Dec_GraphAddNodeAnd(g, na, b), Dec_GraphAddNodeAnd(g, a, nb))
	}
	na, nb := a, b
	na.fCompl ^= 1
	nb.fCompl ^= 1
	e := Dec_GraphAddNodeOr(g, Dec_GraphAddNodeAnd(g, a, b), Dec_GraphAddNodeAnd(g, na, nb))
	e.fCompl ^= 1
	return e
}
func Dec_GraphAddNodeMux(g *Dec_Graph_t, c, t, e Dec_Edge_t, kind int32) Dec_Edge_t {
	if kind != 0 {
		t.fCompl ^= 1
		e.fCompl ^= 1
	}
	nc := c
	nc.fCompl ^= 1
	r := Dec_GraphAddNodeOr(g, Dec_GraphAddNodeAnd(g, c, t), Dec_GraphAddNodeAnd(g, nc, e))
	if kind != 0 {
		r.fCompl ^= 1
	}
	return r
}

// SetLeafSignal binds a graph input to an adapter-owned network signal.
func (g *Dec_Graph_t) SetLeafSignal(index int, signal any) {
	if index < 0 || index >= int(g.nLeaves) {
		panic("dec: leaf index")
	}
	g.pNodes[index].pFunc = signal
}
func (g *Dec_Graph_t) LeafCount() int { return int(g.nLeaves) }
